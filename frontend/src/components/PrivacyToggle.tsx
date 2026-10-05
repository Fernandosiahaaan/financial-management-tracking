import React from 'react';
import { usePrivacy } from '../context/PrivacyContext';

interface PrivacyToggleProps {
  variant?: 'button' | 'icon';
  className?: string;
  id?: string;
}

export const PrivacyToggle: React.FC<PrivacyToggleProps> = ({
  variant = 'button',
  className = '',
  id = 'privacy-toggle-btn',
}) => {
  const { isHidden, togglePrivacy } = usePrivacy();

  const label = isHidden ? 'Show Amounts' : 'Hide Amounts';

  return (
    <button
      type="button"
      className={`privacy-toggle-btn ${variant === 'icon' ? 'privacy-toggle-btn--icon' : ''} ${className}`}
      id={id}
      onClick={togglePrivacy}
      title={label}
      aria-label={label}
      aria-pressed={isHidden}
    >
      <span className="privacy-toggle-icon" aria-hidden="true">
        {isHidden ? (
          <svg
            width="16"
            height="16"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            strokeWidth="2"
            strokeLinecap="round"
            strokeLinejoin="round"
          >
            <path d="M9.88 9.88a3 3 0 1 0 4.24 4.24" />
            <path d="M10.73 5.08A10.43 10.43 0 0 1 12 5c7 0 10 7 10 7a13.16 13.16 0 0 1-1.67 2.68" />
            <path d="M6.61 6.61A13.526 13.526 0 0 0 2 12s3 7 10 7a9.74 9.74 0 0 0 5.39-1.61" />
            <line x1="2" y1="2" x2="22" y2="22" />
          </svg>
        ) : (
          <svg
            width="16"
            height="16"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            strokeWidth="2"
            strokeLinecap="round"
            strokeLinejoin="round"
          >
            <path d="M2 12s3-7 10-7 10 7 10 7-3 7-10 7-10-7-10-7Z" />
            <circle cx="12" cy="12" r="3" />
          </svg>
        )}
      </span>
      {variant === 'button' && (
        <span className="privacy-toggle-label">{isHidden ? 'Show Balance' : 'Hide Balance'}</span>
      )}
    </button>
  );
};
