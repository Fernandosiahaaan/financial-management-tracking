import React, { useState } from 'react';
import { useAuth } from '../context/AuthContext';
import { getPinUserEmail, setRememberedEmail } from '../utils/pinAuth';

interface PinLoginFormProps {
  onSwitchToPassword: () => void;
}

export const PinLoginForm: React.FC<PinLoginFormProps> = ({ onSwitchToPassword }) => {
  const { loginWithPin, isLoading } = useAuth();
  const [pin, setPin] = useState<string>('');
  const [error, setError] = useState<string>('');
  const [isVerifying, setIsVerifying] = useState<boolean>(false);
  const [email, setEmail] = useState<string>(getPinUserEmail());
  const [isEditingEmail, setIsEditingEmail] = useState<boolean>(!getPinUserEmail());

  const handleDigit = async (digit: string) => {
    if (pin.length >= 6 || isVerifying || isLoading) return;
    const newPin = pin + digit;
    setPin(newPin);
    setError('');

    if (newPin.length === 6) {
      await submitPin(newPin);
    }
  };

  const handleBackspace = () => {
    if (isVerifying || isLoading) return;
    setPin((prev) => prev.slice(0, -1));
    setError('');
  };

  const handleClear = () => {
    if (isVerifying || isLoading) return;
    setPin('');
    setError('');
  };

  const submitPin = async (fullPin: string) => {
    if (!email || !email.includes('@')) {
      setError('Masukkan alamat email yang valid terlebih dahulu');
      setPin('');
      setIsEditingEmail(true);
      return;
    }

    setIsVerifying(true);
    setError('');

    try {
      await loginWithPin(email, fullPin);
      setRememberedEmail(email);
    } catch (err) {
      const msg = err instanceof Error ? err.message : 'PIN salah, silakan coba lagi';
      setError(msg);
      setPin('');
      setIsVerifying(false);
    }
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key >= '0' && e.key <= '9') {
      handleDigit(e.key);
    } else if (e.key === 'Backspace') {
      handleBackspace();
    }
  };

  return (
    <div
      className="pin-auth-card auth-card"
      id="pin-login-card"
      onKeyDown={handleKeyDown}
      tabIndex={0}
    >
      <div className="auth-card__header">
        <div className="pin-avatar-badge">👤</div>
        <h2 className="auth-card__title">Quick PIN Login</h2>
        {isEditingEmail ? (
          <div style={{ marginTop: 'var(--space-2)', marginBottom: 'var(--space-2)' }}>
            <input
              type="email"
              placeholder="Masukkan email akun Anda"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              style={{ textAlign: 'center', fontSize: 'var(--font-size-sm)' }}
            />
            {getPinUserEmail() && (
              <button
                type="button"
                className="btn-link"
                style={{ fontSize: '0.8rem', marginTop: 'var(--space-1)' }}
                onClick={() => setIsEditingEmail(false)}
              >
                Gunakan email tersimpan
              </button>
            )}
          </div>
        ) : (
          <p className="auth-card__subtitle">
            Masuk sebagai <strong className="pin-user-highlight">{email}</strong>{' '}
            <button
              type="button"
              className="btn-link"
              style={{ fontSize: '0.8rem', marginLeft: 'var(--space-2)' }}
              onClick={() => setIsEditingEmail(true)}
            >
              (Ganti)
            </button>
          </p>
        )}
      </div>

      {error && (
        <div className="auth-alert auth-alert--error" role="alert" id="pin-error-banner">
          <span className="auth-alert__icon">⚠️</span>
          <span>{error}</span>
        </div>
      )}

      {/* 6-Digit PIN Indicators */}
      <div className="pin-dots-container" id="pin-indicators">
        {[0, 1, 2, 3, 4, 5].map((index) => (
          <div
            key={index}
            className={`pin-dot ${index < pin.length ? 'pin-dot--filled' : ''} ${
              error ? 'pin-dot--error' : ''
            }`}
          />
        ))}
      </div>

      {/* Touch-Friendly Mobile Keypad */}
      <div className="pin-keypad" id="pin-keypad">
        {['1', '2', '3', '4', '5', '6', '7', '8', '9'].map((digit) => (
          <button
            key={digit}
            type="button"
            className="pin-key"
            onClick={() => handleDigit(digit)}
            disabled={isVerifying || isLoading}
          >
            {digit}
          </button>
        ))}
        <button
          type="button"
          className="pin-key pin-key--action"
          onClick={handleClear}
          disabled={isVerifying || isLoading || pin.length === 0}
          aria-label="Hapus semua"
        >
          C
        </button>
        <button
          type="button"
          className="pin-key"
          onClick={() => handleDigit('0')}
          disabled={isVerifying || isLoading}
        >
          0
        </button>
        <button
          type="button"
          className="pin-key pin-key--action"
          onClick={handleBackspace}
          disabled={isVerifying || isLoading || pin.length === 0}
          aria-label="Hapus digit"
        >
          ⌫
        </button>
      </div>

      <div className="pin-footer-actions">
        <button
          type="button"
          className="btn-link pin-switch-btn"
          id="pin-switch-to-password-btn"
          onClick={onSwitchToPassword}
          disabled={isVerifying || isLoading}
        >
          🔑 Masuk dengan Email & Password
        </button>
      </div>
    </div>
  );
};
