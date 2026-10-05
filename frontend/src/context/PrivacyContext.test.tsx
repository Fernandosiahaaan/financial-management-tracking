import { describe, it, expect } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { PrivacyProvider, usePrivacy } from './PrivacyContext';

const TestComponent = () => {
  const { isHidden, togglePrivacy, formatAmount, formatNumber, maskValue } = usePrivacy();

  return (
    <div>
      <span data-testid="status">{isHidden ? 'HIDDEN' : 'VISIBLE'}</span>
      <span data-testid="amount">{formatAmount(1000000000)}</span>
      <span data-testid="neg-amount">{formatAmount(-50000000)}</span>
      <span data-testid="raw-number">{formatNumber(1000000000)}</span>
      <span data-testid="masked-val">{maskValue('Rp 25.000.000,00')}</span>
      <button onClick={togglePrivacy}>Toggle Privacy</button>
    </div>
  );
};

describe('PrivacyContext', () => {
  it('defaults to hidden when rendered within PrivacyProvider', () => {
    render(
      <PrivacyProvider defaultHidden={true}>
        <TestComponent />
      </PrivacyProvider>
    );

    expect(screen.getByTestId('status').textContent).toBe('HIDDEN');
    expect(screen.getByTestId('amount').textContent).toBe('Rp ••••••••');
    expect(screen.getByTestId('neg-amount').textContent).toBe('-Rp ••••••••');
    expect(screen.getByTestId('raw-number').textContent).toBe('••••••••');
    expect(screen.getByTestId('masked-val').textContent).toBe('Rp ••••••••');
  });

  it('toggles between hidden and visible state', () => {
    render(
      <PrivacyProvider defaultHidden={true}>
        <TestComponent />
      </PrivacyProvider>
    );

    const toggleBtn = screen.getByRole('button', { name: /Toggle Privacy/i });

    // Initial state: Hidden
    expect(screen.getByTestId('status').textContent).toBe('HIDDEN');

    // Click to reveal
    fireEvent.click(toggleBtn);
    expect(screen.getByTestId('status').textContent).toBe('VISIBLE');
    expect(screen.getByTestId('amount').textContent).toBe('Rp 10.000.000,00');
    expect(screen.getByTestId('neg-amount').textContent).toBe('-Rp 500.000,00');
    expect(screen.getByTestId('raw-number').textContent).toBe('10.000.000,00');
    expect(screen.getByTestId('masked-val').textContent).toBe('Rp 25.000.000,00');

    // Click to hide again
    fireEvent.click(toggleBtn);
    expect(screen.getByTestId('status').textContent).toBe('HIDDEN');
    expect(screen.getByTestId('amount').textContent).toBe('Rp ••••••••');
  });

  it('falls back gracefully to visible when used outside PrivacyProvider', () => {
    // Tests that standalone components without provider default to unmasked so existing tests pass
    render(<TestComponent />);

    expect(screen.getByTestId('status').textContent).toBe('VISIBLE');
    expect(screen.getByTestId('amount').textContent).toBe('Rp 10.000.000,00');
  });
});
