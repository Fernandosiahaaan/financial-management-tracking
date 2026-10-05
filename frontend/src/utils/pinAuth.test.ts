import { describe, it, expect, beforeEach } from 'vitest';
import {
  hashPin,
  isPinConfigured,
  setupPin,
  verifyPin,
  removePin,
  getPinUserEmail,
} from './pinAuth';

describe('PIN Auth Utility', () => {
  beforeEach(() => {
    localStorage.clear();
  });

  it('correctly hashes a PIN', async () => {
    const hash1 = await hashPin('123456');
    const hash2 = await hashPin('123456');
    const hash3 = await hashPin('654321');

    expect(hash1).toBe(hash2);
    expect(hash1).not.toBe(hash3);
  });

  it('rejects invalid PIN formats during setup', async () => {
    await expect(setupPin('12345', 'user@example.com', 'token-1')).rejects.toThrow();
    await expect(setupPin('1234567', 'user@example.com', 'token-1')).rejects.toThrow();
    await expect(setupPin('abcdef', 'user@example.com', 'token-1')).rejects.toThrow();
  });

  it('stores and verifies valid PIN correctly', async () => {
    expect(isPinConfigured()).toBe(false);

    await setupPin('123456', 'investor@example.com', 'my-auth-token-123');
    expect(isPinConfigured()).toBe(true);
    expect(getPinUserEmail()).toBe('investor@example.com');

    // Correct PIN
    const token = await verifyPin('123456');
    expect(token).toBe('my-auth-token-123');

    // Wrong PIN
    const wrongToken = await verifyPin('000000');
    expect(wrongToken).toBeNull();
  });

  it('removes PIN properly', async () => {
    await setupPin('123456', 'investor@example.com', 'my-token');
    expect(isPinConfigured()).toBe(true);

    removePin();
    expect(isPinConfigured()).toBe(false);
    expect(getPinUserEmail()).toBe('');
  });
});
