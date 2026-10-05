import React, { useState, useEffect } from 'react';
import { useAuth } from '../context/AuthContext';
import * as authApi from '../api/auth';
import {
  isPinConfigured,
  setupPin,
  removePin,
  getPinUserEmail,
} from '../utils/pinAuth';

export const PinSettingsCard: React.FC = () => {
  const { user } = useAuth();
  const [hasPin, setHasPin] = useState<boolean>(false);
  const [showSetupModal, setShowSetupModal] = useState<boolean>(false);
  const [pinInput, setPinInput] = useState<string>('');
  const [confirmPinInput, setConfirmPinInput] = useState<string>('');
  const [error, setError] = useState<string>('');
  const [success, setSuccess] = useState<string>('');

  useEffect(() => {
    setHasPin(isPinConfigured());
  }, []);

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

    const token = authApi.getToken();
    if (!token || !user?.email) {
      setError('Gagal membaca sesi autentikasi. Silakan refresh halaman.');
      return;
    }

    try {
      await setupPin(pinInput, user.email, token);
      setHasPin(true);
      setShowSetupModal(false);
      setSuccess('Quick PIN 6-digit berhasil diaktifkan di perangkat ini!');
      setTimeout(() => setSuccess(''), 4000);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Gagal menyimpan PIN');
    }
  };

  const handleRemovePin = () => {
    if (window.confirm('Apakah Anda yakin ingin menonaktifkan Quick PIN di perangkat ini?')) {
      removePin();
      setHasPin(false);
      setSuccess('Quick PIN telah dinonaktifkan.');
      setTimeout(() => setSuccess(''), 3000);
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
          <span className="pin-status-label">Status Quick PIN di Perangkat Ini:</span>
          <span className={`badge ${hasPin ? 'badge--success' : 'badge--neutral'}`}>
            {hasPin ? '🟢 Aktif' : '⚪ Belum Diatur'}
          </span>
          {hasPin && (
            <span className="pin-status-user">
              Terkait dengan: <strong>{getPinUserEmail()}</strong>
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
              >
                Ubah PIN
              </button>
              <button
                type="button"
                className="btn btn-danger-outline btn--sm"
                onClick={handleRemovePin}
                id="pin-remove-btn"
              >
                Hapus PIN
              </button>
            </>
          ) : (
            <button
              type="button"
              className="btn btn--primary btn--sm"
              onClick={handleOpenSetup}
              id="pin-setup-btn"
            >
              🔒 Aktifkan Quick PIN
            </button>
          )}
        </div>
      </div>

      {/* Setup PIN Modal */}
      {showSetupModal && (
        <div className="modal-overlay" onClick={() => setShowSetupModal(false)}>
          <div className="modal-content" onClick={(e) => e.stopPropagation()} style={{ maxWidth: '400px' }}>
            <div className="modal-header">
              <h3>{hasPin ? 'Ubah Quick PIN' : 'Atur Quick PIN Baru'}</h3>
              <button className="btn-close" onClick={() => setShowSetupModal(false)}>✕</button>
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
                  required
                />
              </div>

              <div className="modal-actions">
                <button
                  type="button"
                  className="btn btn--outline"
                  onClick={() => setShowSetupModal(false)}
                >
                  Batal
                </button>
                <button
                  type="submit"
                  className="btn btn--primary"
                  disabled={pinInput.length !== 6 || confirmPinInput.length !== 6}
                >
                  Simpan PIN
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
};
