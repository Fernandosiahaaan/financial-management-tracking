import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import * as authApi from './auth';

describe('Auth API client', () => {
  const originalFetch = globalThis.fetch;

  beforeEach(() => {
    vi.restoreAllMocks();
    localStorage.clear();
  });

  afterEach(() => {
    globalThis.fetch = originalFetch;
  });

  it('register saves token to localStorage and returns AuthResponseData', async () => {
    const mockData = {
      token: 'mock-jwt-token-123',
      profile: {
        id: 'user-uuid-1',
        email: 'test@example.com',
        cycle_start_day: 15,
        created_at: '2026-10-04T00:00:00Z',
      },
    };

    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: true,
      status: 201,
      json: async () => ({ success: true, data: mockData }),
    });

    const result = await authApi.register('test@example.com', 'Pass1234!', 15);
    expect(result).toEqual(mockData);
    expect(localStorage.getItem('fintrack_token')).toBe('mock-jwt-token-123');
  });

  it('login stores token and returns user profile', async () => {
    const mockData = {
      token: 'logged-in-jwt',
      profile: {
        id: 'user-uuid-2',
        email: 'login@example.com',
        cycle_start_day: 1,
        created_at: '2026-10-04T00:00:00Z',
      },
    };

    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: true,
      status: 200,
      json: async () => ({ success: true, data: mockData }),
    });

    const result = await authApi.login('login@example.com', 'Pass1234!');
    expect(result).toEqual(mockData);
    expect(authApi.getToken()).toBe('logged-in-jwt');
  });

  it('logout removes token from localStorage', async () => {
    authApi.setToken('existing-token');
    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: true,
      status: 200,
      json: async () => ({ success: true }),
    });

    await authApi.logout();
    expect(authApi.getToken()).toBeNull();
  });

  it('getMe sends Bearer token and returns profile', async () => {
    authApi.setToken('my-auth-token');
    const mockProfile = {
      id: 'user-uuid-3',
      email: 'profile@example.com',
      cycle_start_day: 1,
      created_at: '2026-10-04T00:00:00Z',
    };

    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: true,
      status: 200,
      json: async () => ({ success: true, data: mockProfile }),
    });

    const profile = await authApi.getMe();
    expect(profile).toEqual(mockProfile);
    expect(globalThis.fetch).toHaveBeenCalledWith(
      expect.stringContaining('/api/v1/auth/me'),
      expect.objectContaining({
        headers: { Authorization: 'Bearer my-auth-token' },
      })
    );
  });

  it('updateUserSettings sends PUT with updated cycle_start_day', async () => {
    authApi.setToken('my-auth-token');
    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: true,
      status: 200,
      json: async () => ({ success: true, data: { cycle_start_day: 28 } }),
    });

    const res = await authApi.updateUserSettings(28);
    expect(res.cycle_start_day).toBe(28);
  });
});
