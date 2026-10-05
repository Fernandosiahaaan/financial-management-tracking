import React, { useState } from 'react';
import { useAuth } from '../context/AuthContext';

export const UserSettingsCard: React.FC = () => {
  const { user, updateCycleStartDay } = useAuth();
  const [cycleDay, setCycleDay] = useState<number>(user?.cycle_start_day || 1);
  const [isSaving, setIsSaving] = useState<boolean>(false);
  const [successMessage, setSuccessMessage] = useState<string>('');
  const [errorMessage, setErrorMessage] = useState<string>('');

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

  return (
    <div className="settings-card" id="user-settings-card">
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
  );
};
