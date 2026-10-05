import React, { useState } from 'react';
import { useAuth } from '../context/AuthContext';
import * as authApi from '../api/auth';
import { setRememberedEmail, clearRememberedEmail } from '../utils/pinAuth';

export const PinSettingsCard: React.FC = () => {
  const { user, refreshProfile } = useAuth();
  const [showSetupModal, setShowSetupModal] = useState<boolean>(false);
  const [pinInput, setPinInput] = useState<string>('');
  const [confirmPinInput, setConfirmPinInput] = useState<string>('');
  const [error, setError] = useState<string>('');
  const [success, setSuccess] = useState<string>('');
  const [isSaving, setIsSaving] = useState<boolean>(false);

  const hasPin = Boolean(user?.has_pin);

  const handleOpenSetup = () => {
    setPinInput('');
    setConfirmPinInput('');
    setError('');
    setShowSetupModal(true);
  };

  const handleSavePin = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');

    if (pinInput.length !== 6 || !/^\d{6}$/.test(pinInput)) {
      setError('PIN harus terdiri dari 6 digit angka');
      return;
    }

    if (pinInput !== confirmPinInput) {
      setError('Konfirmasi PIN tidak cocok');
      return;
    }

    setIsSaving(true);
    try {
      await authApi.setPin(pinInput);
      if (user?.email) {
        setRememberedEmail(user.email);
      }
      await refreshProfile();
      setShowSetupModal(false);
      setSuccess('PIN 6-digit berhasil disimpan ke database akun Anda! Anda sekarang dapat login menggunakan PIN di perangkat mana pun.');
      setTimeout(() => setSuccess(''), 5000);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Gagal menyimpan PIN ke server');
    } finally {
      setIsSaving(false);
    }
  };

  const handleRemovePin = async () => {
    if (window.confirm('Apakah Anda yakin ingin menonaktifkan PIN untuk akun Anda?')) {
      setIsSaving(true);
      try {
        await authApi.removePin();
        clearRememberedEmail();
        await refreshProfile();
        setSuccess('PIN telah berhasil dinonaktifkan dari database akun Anda.');
        setTimeout(() => setSuccess(''), 4000);
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Gagal menonaktifkan PIN');
      } finally {
        setIsSaving(false);
      }
    }
  };

  return (
    <div className="settings-card" id="pin-settings-card" style={{ marginTop: 'var(--space-4)' }}>
      <div className="settings-card__header">
        <h3 className="settings-card__title">📱 Quick PIN Login (Akses Cepat di HP)</h3>
        <p className="settings-card__subtitle">
          Gunakan 6-digit PIN untuk login instan tanpa perlu mengetik email & password berulang kali di HP/browser ini.
        </p>
      </div>

      {success && (
        <div className="auth-alert auth-alert--success" role="status" id="pin-settings-success">
          <span className="auth-alert__icon">✅</span>
          <span>{success}</span>
        </div>
      )}

      <div className="pin-settings-status-box">
        <div className="pin-status-info">
          <span className="pin-status-label">Status PIN Akun:</span>
          <span className={`badge ${hasPin ? 'badge--success' : 'badge--neutral'}`}>
            {hasPin ? '🟢 Aktif di Database' : '⚪ Belum Diatur'}
          </span>
          {hasPin && (
            <span className="pin-status-user">
              Tersimpan untuk akun: <strong>{user?.email}</strong>
            </span>
          )}
        </div>

        <div className="pin-status-actions">
          {hasPin ? (
            <>
              <button
                type="button"
                className="btn btn--outline btn--sm"
                onClick={handleOpenSetup}
                id="pin-change-btn"
                disabled={isSaving}
              >
                Ubah PIN
              </button>
              <button
                type="button"
                className="btn btn-danger-outline btn--sm"
                onClick={handleRemovePin}
                id="pin-remove-btn"
                disabled={isSaving}
              >
                {isSaving ? 'Menghapus…' : 'Hapus PIN'}
              </button>
            </>
          ) : (
            <button
              type="button"
              className="btn btn--primary btn--sm"
              onClick={handleOpenSetup}
              id="pin-setup-btn"
              disabled={isSaving}
            >
              🔒 Aktifkan PIN
            </button>
          )}
        </div>
      </div>

      {/* Setup PIN Modal */}
      {showSetupModal && (
        <div className="modal-overlay" onClick={() => !isSaving && setShowSetupModal(false)}>
          <div className="modal-content" onClick={(e) => e.stopPropagation()} style={{ maxWidth: '400px' }}>
            <div className="modal-header">
              <h3>{hasPin ? 'Ubah PIN Akun' : 'Atur PIN Akun Baru'}</h3>
              <button className="btn-close" onClick={() => !isSaving && setShowSetupModal(false)}>✕</button>
            </div>

            {error && (
              <div className="auth-alert auth-alert--error" style={{ marginBottom: 'var(--space-3)' }}>
                <span>⚠️ {error}</span>
              </div>
            )}

            <form onSubmit={handleSavePin}>
              <div className="form-group" style={{ marginBottom: 'var(--space-3)' }}>
                <label htmlFor="new-pin-input">Masukkan 6-Digit PIN Baru</label>
                <input
                  id="new-pin-input"
                  type="password"
                  inputMode="numeric"
                  pattern="[0-9]*"
                  maxLength={6}
                  placeholder="Contoh: 123456"
                  value={pinInput}
                  onChange={(e) => setPinInput(e.target.value.replace(/\D/g, '').slice(0, 6))}
                  autoFocus
                  disabled={isSaving}
                  required
                />
              </div>

              <div className="form-group" style={{ marginBottom: 'var(--space-4)' }}>
                <label htmlFor="confirm-pin-input">Konfirmasi 6-Digit PIN</label>
                <input
                  id="confirm-pin-input"
                  type="password"
                  inputMode="numeric"
                  pattern="[0-9]*"
                  maxLength={6}
                  placeholder="Ketik ulang 6 digit PIN"
                  value={confirmPinInput}
                  onChange={(e) => setConfirmPinInput(e.target.value.replace(/\D/g, '').slice(0, 6))}
                  disabled={isSaving}
                  required
                />
              </div>

              <div className="modal-actions">
                <button
                  type="button"
                  className="btn btn--outline"
                  onClick={() => setShowSetupModal(false)}
                  disabled={isSaving}
                >
                  Batal
                </button>
                <button
                  type="submit"
                  className="btn btn--primary"
                  disabled={isSaving || pinInput.length !== 6 || confirmPinInput.length !== 6}
                >
                  {isSaving ? 'Menyimpan…' : 'Simpan PIN'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
};
