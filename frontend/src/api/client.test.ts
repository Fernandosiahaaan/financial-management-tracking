import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { fetchHealth } from './client';

describe('fetchHealth API client', () => {
  const originalFetch = globalThis.fetch;

  beforeEach(() => {
    vi.restoreAllMocks();
  });

  afterEach(() => {
    globalThis.fetch = originalFetch;
  });

  it('returns parsed health response on 200 OK', async () => {
    const mockData = {
      success: true,
      status: 'healthy',
      database: 'connected',
    };

    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: true,
      status: 200,
      json: async () => mockData,
    });

    const result = await fetchHealth();
    expect(result).toEqual(mockData);
    expect(globalThis.fetch).toHaveBeenCalledWith('http://localhost:8080/api/v1/health');
  });

  it('throws an error when response is not ok', async () => {
    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: false,
      status: 503,
      json: async () => ({ success: false, status: 'unhealthy', database: 'disconnected' }),
    });

    await expect(fetchHealth()).rejects.toThrow('Health check failed: 503');
  });

  it('propagates network errors', async () => {
    globalThis.fetch = vi.fn().mockRejectedValue(new Error('Network failure'));

    await expect(fetchHealth()).rejects.toThrow('Network failure');
  });
});
