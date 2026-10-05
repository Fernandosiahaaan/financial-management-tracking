import { describe, it, expect } from 'vitest';
import { API_BASE_URL, getApiUrl, config, normalizeBaseUrl } from './config';

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

  describe('normalizeBaseUrl', () => {
    it('prepends https:// when domain has no protocol', () => {
      expect(normalizeBaseUrl('financial-management-tracking-production.up.railway.app')).toBe(
        'https://financial-management-tracking-production.up.railway.app'
      );
      expect(normalizeBaseUrl('api.myproject.com/')).toBe('https://api.myproject.com');
    });

    it('keeps existing https:// or http:// protocol', () => {
      expect(normalizeBaseUrl('https://api.railway.app')).toBe('https://api.railway.app');
      expect(normalizeBaseUrl('http://custom-domain.com:8080/')).toBe('http://custom-domain.com:8080');
    });

    it('defaults localhost or 127.0.0.1 to http://', () => {
      expect(normalizeBaseUrl('localhost:8080')).toBe('http://localhost:8080');
      expect(normalizeBaseUrl('127.0.0.1:8080')).toBe('http://127.0.0.1:8080');
    });

    it('handles empty or undefined input', () => {
      expect(normalizeBaseUrl(undefined)).toBe('http://localhost:8080');
      expect(normalizeBaseUrl('')).toBe('http://localhost:8080');
      expect(normalizeBaseUrl('   ')).toBe('http://localhost:8080');
    });
  });
});
