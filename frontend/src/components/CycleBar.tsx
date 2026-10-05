import React, { useState, useEffect } from 'react';
import { type CycleSummary, getCycleSummary } from '../api/cycle';
import { usePrivacy } from '../context/PrivacyContext';
import { PrivacyToggle } from './PrivacyToggle';

interface CycleBarProps {
  onCycleChange?: (startDate: string, endDate: string) => void;
  refreshTrigger?: number;
}

export const CycleBar: React.FC<CycleBarProps> = ({ onCycleChange, refreshTrigger }) => {
  const { formatAmount } = usePrivacy();
  const [summary, setSummary] = useState<CycleSummary | null>(null);
  const [targetDate, setTargetDate] = useState<string>('');
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);

  const fetchCycle = async (dateStr?: string) => {
    try {
      setLoading(true);
      setError(null);
      const data = await getCycleSummary(dateStr);
      setSummary(data);
      if (onCycleChange) {
        onCycleChange(data.cycle.start_date, data.cycle.end_date);
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load cycle data');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchCycle(targetDate);
  }, [targetDate, refreshTrigger]);

  const handlePrevCycle = () => {
    if (!summary) return;
    // Go to the day before current cycle start
    const start = new Date(summary.cycle.start_date);
    start.setDate(start.getDate() - 1);
    const dateStr = start.toISOString().split('T')[0];
    setTargetDate(dateStr);
  };

  const handleNextCycle = () => {
    if (!summary) return;
    // Go to the day after current cycle end
    const end = new Date(summary.cycle.end_date);
    end.setDate(end.getDate() + 1);
    const dateStr = end.toISOString().split('T')[0];
    setTargetDate(dateStr);
  };

  const handleResetCurrent = () => {
    setTargetDate('');
  };

  const formatDateHuman = (dateStr: string) => {
    const months = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec'];
    const parts = dateStr.split('-');
    if (parts.length === 3) {
      const year = parts[0];
      const monthIdx = parseInt(parts[1], 10) - 1;
      const day = parseInt(parts[2], 10);
      if (monthIdx >= 0 && monthIdx < 12) {
        return `${day} ${months[monthIdx]} ${year}`;
      }
    }
    return dateStr;
  };

  const percentElapsed = summary
    ? Math.min(100, Math.max(0, Math.round((summary.cycle.day_of_cycle / summary.cycle.total_days) * 100)))
    : 0;

  return (
    <div className="cycle-bar-container" id="cycle-bar">
      <div className="cycle-bar-header">
        <div className="cycle-identity">
          <span className="cycle-icon">🗓️</span>
          <div>
            <div className="cycle-title-row">
              <h3 className="cycle-title">
                {summary
                  ? `${formatDateHuman(summary.cycle.start_date)} — ${formatDateHuman(summary.cycle.end_date)}`
                  : 'Financial Cycle'}
              </h3>
              {summary && (
                <span className="badge badge--cycle-day">
                  Starts Day {summary.cycle.cycle_start_day}
                </span>
              )}
              <PrivacyToggle variant="icon" id="cycle-privacy-btn" />
            </div>
            {summary && (
              <p className="cycle-subtitle">
                Day {summary.cycle.day_of_cycle} of {summary.cycle.total_days} •{' '}
                {summary.cycle.days_remaining} {summary.cycle.days_remaining === 1 ? 'day' : 'days'} remaining
              </p>
            )}
          </div>
        </div>

        <div className="cycle-nav-controls">
          <button
            onClick={handlePrevCycle}
            className="btn-cycle-nav"
            title="Previous Financial Cycle"
            type="button"
            disabled={loading}
          >
            ◀ Prev Cycle
          </button>
          {targetDate && (
            <button
              onClick={handleResetCurrent}
              className="btn-cycle-nav today"
              title="Return to Current Cycle"
              type="button"
              disabled={loading}
            >
              Current Cycle
            </button>
          )}
          <button
            onClick={handleNextCycle}
            className="btn-cycle-nav"
            title="Next Financial Cycle"
            type="button"
            disabled={loading}
          >
            Next Cycle ▶
          </button>
        </div>
      </div>

      {error && <div className="cycle-error-alert">{error}</div>}

      {/* Progress Bar */}
      <div className="cycle-progress-track">
        <div
          className="cycle-progress-fill"
          style={{ width: `${percentElapsed}%` }}
          title={`${percentElapsed}% elapsed`}
        />
      </div>

      {/* Mini Cycle Metrics */}
      {summary && (
        <div className="cycle-metrics-grid">
          <div className="cycle-metric-chip income">
            <span className="chip-label">Cycle Income</span>
            <span className="chip-value">+{formatAmount(summary.total_income)}</span>
          </div>
          <div className="cycle-metric-chip expense">
            <span className="chip-label">Cycle Expense</span>
            <span className="chip-value">-{formatAmount(summary.total_expense)}</span>
          </div>
          <div className={`cycle-metric-chip ${summary.net_savings >= 0 ? 'savings-pos' : 'savings-neg'}`}>
            <span className="chip-label">Net Savings</span>
            <span className="chip-value">
              {summary.net_savings >= 0 ? '+' : ''}
              {formatAmount(summary.net_savings)}
            </span>
          </div>
        </div>
      )}
    </div>
  );
};
