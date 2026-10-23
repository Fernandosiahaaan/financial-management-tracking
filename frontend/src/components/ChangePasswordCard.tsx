import React, { useState } from 'react';
import { changePassword } from '../api/auth';
import { KeyRound, Eye, EyeOff, CheckCircle2, AlertCircle } from 'lucide-react';

export const ChangePasswordCard: React.FC = () => {
  const [currentPassword, setCurrentPassword] = useState('');
  const [newPassword, setNewPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');

  const [showCurrent, setShowCurrent] = useState(false);
  const [showNew, setShowNew] = useState(false);
  const [showConfirm, setShowConfirm] = useState(false);

  const [isLoading, setIsLoading] = useState(false);
  const [successMessage, setSuccessMessage] = useState('');
  const [errorMessage, setErrorMessage] = useState('');

  const resetForm = () => {
    setCurrentPassword('');
    setNewPassword('');
    setConfirmPassword('');
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setSuccessMessage('');
    setErrorMessage('');

    if (!currentPassword) {
      setErrorMessage('Kata sandi saat ini harus diisi');
      return;
    }
    if (newPassword.length < 8) {
      setErrorMessage('Kata sandi baru minimal 8 karakter');
      return;
    }
    if (newPassword === currentPassword) {
      setErrorMessage('Kata sandi baru tidak boleh sama dengan kata sandi saat ini');
      return;
    }
    if (newPassword !== confirmPassword) {
      setErrorMessage('Konfirmasi kata sandi baru tidak cocok');
      return;
    }

    setIsLoading(true);
    try {
      await changePassword(currentPassword, newPassword);
      setSuccessMessage('Kata sandi berhasil diperbarui!');
      resetForm();
      setTimeout(() => setSuccessMessage(''), 5000);
    } catch (err) {
      const msg = err instanceof Error ? err.message : 'Gagal memperbarui kata sandi';
      setErrorMessage(msg);
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="settings-card" id="change-password-card" style={{ marginTop: 'var(--space-4)' }}>
      <div className="settings-card__header">
        <div style={{ display: 'flex', alignItems: 'center', gap: '0.625rem' }}>
          <div className="brand-icon-box" style={{ width: '32px', height: '32px', background: 'rgba(99, 102, 241, 0.15)', color: 'var(--color-brand-primary)' }}>
            <KeyRound size={17} />
          </div>
          <div>
            <h3 className="settings-card__title">Ganti Kata Sandi</h3>
            <p className="settings-card__subtitle">
              Perbarui kata sandi akun FinTrack Anda secara aman untuk menjaga keamanan data finansial.
            </p>
          </div>
        </div>
      </div>

      {successMessage && (
        <div className="auth-alert auth-alert--success" role="status" id="change-password-success-banner" style={{ marginTop: 'var(--space-3)' }}>
          <CheckCircle2 size={18} className="auth-alert__icon" />
          <span>{successMessage}</span>
        </div>
      )}

      {errorMessage && (
        <div className="auth-alert auth-alert--error" role="alert" id="change-password-error-banner" style={{ marginTop: 'var(--space-3)' }}>
          <AlertCircle size={18} className="auth-alert__icon" />
          <span>{errorMessage}</span>
        </div>
      )}

      <form className="settings-form" onSubmit={handleSubmit} style={{ marginTop: 'var(--space-4)' }}>
        <div className="form-group">
          <label htmlFor="current-password">Kata Sandi Saat Ini</label>
          <div className="password-input-wrapper">
            <input
              id="current-password"
              type={showCurrent ? 'text' : 'password'}
              placeholder="Masukkan kata sandi lama"
              value={currentPassword}
              onChange={(e) => {
                setCurrentPassword(e.target.value);
                setErrorMessage('');
              }}
              disabled={isLoading}
              required
            />
            <button
              type="button"
              className="password-toggle-btn"
              id="toggle-current-password"
              onClick={() => setShowCurrent((prev) => !prev)}
              aria-label={showCurrent ? 'Sembunyikan kata sandi' : 'Tampilkan kata sandi'}
              tabIndex={-1}
            >
              {showCurrent ? <EyeOff size={18} /> : <Eye size={18} />}
            </button>
          </div>
        </div>

        <div className="form-row">
          <div className="form-group flex-1">
            <label htmlFor="new-password">Kata Sandi Baru</label>
            <div className="password-input-wrapper">
              <input
                id="new-password"
                type={showNew ? 'text' : 'password'}
                placeholder="Minimal 8 karakter"
                value={newPassword}
                onChange={(e) => {
                  setNewPassword(e.target.value);
                  setErrorMessage('');
                }}
                disabled={isLoading}
                required
              />
              <button
                type="button"
                className="password-toggle-btn"
                id="toggle-new-password"
                onClick={() => setShowNew((prev) => !prev)}
                aria-label={showNew ? 'Sembunyikan kata sandi' : 'Tampilkan kata sandi'}
                tabIndex={-1}
              >
                {showNew ? <EyeOff size={18} /> : <Eye size={18} />}
              </button>
            </div>
            <span className="form-hint">Gunakan kombinasi huruf, angka, dan simbol.</span>
          </div>

          <div className="form-group flex-1">
            <label htmlFor="confirm-new-password">Konfirmasi Kata Sandi Baru</label>
            <div className="password-input-wrapper">
              <input
                id="confirm-new-password"
                type={showConfirm ? 'text' : 'password'}
                placeholder="Ulangi kata sandi baru"
                value={confirmPassword}
                onChange={(e) => {
                  setConfirmPassword(e.target.value);
                  setErrorMessage('');
                }}
                disabled={isLoading}
                required
              />
              <button
                type="button"
                className="password-toggle-btn"
                id="toggle-confirm-password"
                onClick={() => setShowConfirm((prev) => !prev)}
                aria-label={showConfirm ? 'Sembunyikan kata sandi' : 'Tampilkan kata sandi'}
                tabIndex={-1}
              >
                {showConfirm ? <EyeOff size={18} /> : <Eye size={18} />}
              </button>
            </div>
          </div>
        </div>

        <div className="form-actions" style={{ marginTop: 'var(--space-2)' }}>
          <button
            type="submit"
            className="btn btn--primary"
            id="btn-update-password"
            disabled={isLoading || !currentPassword || !newPassword || !confirmPassword}
          >
            {isLoading ? 'Menyimpan…' : 'Perbarui Kata Sandi'}
          </button>
        </div>
      </form>
    </div>
  );
};
