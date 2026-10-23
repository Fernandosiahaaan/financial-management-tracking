import React, { useState, useEffect } from 'react';
import { useAuth } from '../context/AuthContext';
import { PinLoginForm } from './PinLoginForm';
import { isPinConfigured } from '../utils/pinAuth';
import { Eye, EyeOff } from 'lucide-react';

export const AuthForms: React.FC = () => {
  const { login, register, error, clearError, isLoading } = useAuth();
  const [isRegister, setIsRegister] = useState(false);
  const [usePinMode, setUsePinMode] = useState(false);
  const [hasPinOnDevice, setHasPinOnDevice] = useState(false);
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [showPassword, setShowPassword] = useState(false);
  const [cycleStartDay, setCycleStartDay] = useState(1);
  const [clientError, setClientError] = useState('');

  useEffect(() => {
    setHasPinOnDevice(isPinConfigured());
  }, []);

  const validate = (): boolean => {
    if (!email || !email.includes('@')) {
      setClientError('Please enter a valid email address');
      return false;
    }
    if (password.length < 8) {
      setClientError('Password must be at least 8 characters long');
      return false;
    }
    if (isRegister && (cycleStartDay < 1 || cycleStartDay > 31)) {
      setClientError('Cycle start day must be between 1 and 31');
      return false;
    }
    setClientError('');
    return true;
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    clearError();
    if (!validate()) return;

    try {
      if (isRegister) {
        await register(email, password, cycleStartDay);
      } else {
        await login(email, password);
      }
    } catch {
      // Errors handled via context
    }
  };

  const toggleMode = () => {
    setIsRegister(!isRegister);
    setClientError('');
    clearError();
  };

  const displayedError = clientError || error;

  if (usePinMode && hasPinOnDevice) {
    return <PinLoginForm onSwitchToPassword={() => setUsePinMode(false)} />;
  }

  return (
    <div className="auth-card" id="auth-card">
      {hasPinOnDevice && !isRegister && (
        <button
          type="button"
          className="btn btn--outline btn--full pin-quick-login-toggle"
          id="auth-use-pin-btn"
          onClick={() => setUsePinMode(true)}
          style={{ marginBottom: 'var(--space-4)' }}
        >
          🔒 Masuk dengan Quick PIN
        </button>
      )}

      <div className="auth-card__header">
        <h2 className="auth-card__title">
          {isRegister ? 'Create your account' : 'Welcome back'}
        </h2>
        <p className="auth-card__subtitle">
          {isRegister
            ? 'Start managing your finances with precision'
            : 'Enter your credentials to access FinTrack'}
        </p>
      </div>

      {displayedError && (
        <div className="auth-alert auth-alert--error" role="alert" id="auth-error-banner">
          <span className="auth-alert__icon">⚠️</span>
          <span>{displayedError}</span>
        </div>
      )}

      <form className="auth-form" onSubmit={handleSubmit} noValidate>
        <div className="form-group">
          <label htmlFor="auth-email">Email Address</label>
          <input
            id="auth-email"
            type="email"
            placeholder="you@example.com"
            value={email}
            onChange={(e) => {
              setEmail(e.target.value);
              setClientError('');
            }}
            disabled={isLoading}
            required
          />
        </div>

        <div className="form-group">
          <label htmlFor="auth-password">Password</label>
          <div className="password-input-wrapper">
            <input
              id="auth-password"
              type={showPassword ? 'text' : 'password'}
              placeholder="••••••••"
              value={password}
              onChange={(e) => {
                setPassword(e.target.value);
                setClientError('');
              }}
              disabled={isLoading}
              required
            />
            <button
              type="button"
              className="password-toggle-btn"
              id="auth-password-toggle-btn"
              onClick={() => setShowPassword((prev) => !prev)}
              aria-label={showPassword ? 'Hide password' : 'Show password'}
              title={showPassword ? 'Hide password' : 'Show password'}
              tabIndex={-1}
            >
              {showPassword ? <EyeOff size={18} /> : <Eye size={18} />}
            </button>
          </div>
          <span className="form-hint">At least 8 characters</span>
        </div>

        {isRegister && (
          <div className="form-group">
            <label htmlFor="auth-cycle-day">Financial Cycle Start Day</label>
            <input
              id="auth-cycle-day"
              type="number"
              min={1}
              max={31}
              value={cycleStartDay}
              onChange={(e) => setCycleStartDay(parseInt(e.target.value, 10) || 1)}
              disabled={isLoading}
            />
            <span className="form-hint">e.g. 25 for salary payday</span>
          </div>
        )}

        <button
          type="submit"
          className="btn btn--primary btn--full"
          id="auth-submit-btn"
          disabled={isLoading}
        >
          {isLoading ? (
            <span className="btn-spinner">Processing…</span>
          ) : isRegister ? (
            'Create Account'
          ) : (
            'Sign In'
          )}
        </button>
      </form>

      <div className="auth-card__footer">
        <span>
          {isRegister ? 'Already have an account?' : "Don't have an account?"}
        </span>
        <button
          type="button"
          className="btn-link"
          id="auth-toggle-mode-btn"
          onClick={toggleMode}
          disabled={isLoading}
        >
          {isRegister ? 'Sign in instead' : 'Create an account'}
        </button>
      </div>
    </div>
  );
};
