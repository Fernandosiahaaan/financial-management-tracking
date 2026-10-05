import { describe, it, expect } from 'vitest';
import { API_BASE_URL, getApiUrl, config } from './config';

describe('Frontend Config', () => {
  it('provides a valid API_BASE_URL', () => {
    expect(typeof API_BASE_URL).toBe('string');
    expect(config.api.baseUrl).toBe(API_BASE_URL);
  });

  it('getApiUrl appends relative paths properly', () => {
    const url = getApiUrl('/api/v1/health');
    expect(url).toContain('/api/v1/health');
    expect(config.api.getUrl('/api/v1/accounts')).toContain('/api/v1/accounts');
  });

  it('getApiUrl handles paths without leading slash', () => {
    const url = getApiUrl('api/v1/health');
    expect(url).toContain('/api/v1/health');
  });
});
