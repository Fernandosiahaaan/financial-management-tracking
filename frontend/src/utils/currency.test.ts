import { describe, it, expect } from 'vitest';
import { formatRupiah, formatNumberIDR, parseRupiahToCents, formatCurrencyInput } from './currency';

describe('currency utilities', () => {
  describe('formatRupiah', () => {
    it('formats cents into standard IDR currency with Rp prefix and ,00', () => {
      expect(formatRupiah(10000000)).toBe('Rp 100.000,00');
      expect(formatRupiah(10000050)).toBe('Rp 100.000,50');
      expect(formatRupiah(500000)).toBe('Rp 5.000,00');
      expect(formatRupiah(0)).toBe('Rp 0,00');
      expect(formatRupiah(-500000)).toBe('-Rp 5.000,00');
    });
  });

  describe('formatNumberIDR', () => {
    it('formats cents into xxx.xxx,00 format without prefix', () => {
      expect(formatNumberIDR(10000000)).toBe('100.000,00');
      expect(formatNumberIDR(10000050)).toBe('100.000,50');
      expect(formatNumberIDR(500000)).toBe('5.000,00');
      expect(formatNumberIDR(0)).toBe('0,00');
      expect(formatNumberIDR(-500000)).toBe('-5.000,00');
    });
  });

  describe('parseRupiahToCents', () => {
    it('correctly parses various Indonesian string formats into cents', () => {
      expect(parseRupiahToCents('100.000,00')).toBe(10000000);
      expect(parseRupiahToCents('100.000,50')).toBe(10000050);
      expect(parseRupiahToCents('100.000')).toBe(10000000);
      expect(parseRupiahToCents('Rp 100.000,00')).toBe(10000000);
      expect(parseRupiahToCents('100000')).toBe(10000000);
      expect(parseRupiahToCents('100000.50')).toBe(10000050);
      expect(parseRupiahToCents('0')).toBe(0);
      expect(parseRupiahToCents('')).toBe(0);
    });
  });

  describe('formatCurrencyInput', () => {
    it('formats user input into xxx.xxx,00', () => {
      expect(formatCurrencyInput('100000')).toBe('100.000,00');
      expect(formatCurrencyInput('100.000')).toBe('100.000,00');
      expect(formatCurrencyInput('100000,50')).toBe('100.000,50');
    });
  });
});
