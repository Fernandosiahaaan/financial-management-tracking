/**
 * Currency utility for Indonesian Rupiah formatting and parsing.
 * Backend stores amounts as integer Cents / Sen (1 IDR = 100 Sen).
 * Frontend formats display and inputs as 'Rp xxx.xxx,00' or 'xxx.xxx,00'.
 */

/**
 * Formats a monetary amount in cents into standard Indonesian currency string:
 * e.g. 10000000 cents -> "Rp 100.000,00"
 * e.g. 10000050 cents -> "Rp 100.000,50"
 * e.g. -500000 cents  -> "-Rp 5.000,00"
 */
export function formatRupiah(amountInCents: number): string {
  const isNegative = amountInCents < 0;
  const absCents = Math.abs(amountInCents);
  const formattedNumber = formatNumberIDR(absCents);
  return isNegative ? `-Rp ${formattedNumber}` : `Rp ${formattedNumber}`;
}

/**
 * Formats a monetary amount in cents into "xxx.xxx,00" format without currency prefix:
 * e.g. 10000000 cents -> "100.000,00"
 * e.g. 500000 cents   -> "5.000,00"
 * e.g. 0 cents        -> "0,00"
 */
export function formatNumberIDR(amountInCents: number): string {
  const isNegative = amountInCents < 0;
  const absCents = Math.abs(Math.round(amountInCents));
  const whole = Math.floor(absCents / 100);
  const dec = absCents % 100;

  // Format whole part with thousand dots
  const wholeStr = whole.toString().replace(/\B(?=(\d{3})+(?!\d))/g, '.');
  const decStr = dec.toString().padStart(2, '0');

  const res = `${wholeStr},${decStr}`;
  return isNegative ? `-${res}` : res;
}

/**
 * Parses user input string (which can be "100.000,00", "100.000", "100000", "Rp 100.000,00", etc.)
 * into integer cents / sen.
 */
export function parseRupiahToCents(val: string | number): number {
  if (typeof val === 'number') {
    return Math.round(val);
  }

  let clean = val.trim();
  if (!clean) return 0;

  // Strip "Rp", "rp", spaces
  clean = clean.replace(/^[Rr][Pp]\.?\s*/, '').trim();

  const isNegative = clean.startsWith('-');
  if (isNegative) {
    clean = clean.substring(1).trim();
  }

  let wholePart = '';
  let decimalPart = '00';

  if (clean.includes(',')) {
    const parts = clean.split(',');
    wholePart = parts[0].replace(/\D/g, '');
    decimalPart = parts[1].replace(/\D/g, '').padEnd(2, '0').slice(0, 2);
  } else if (clean.includes('.')) {
    const dotParts = clean.split('.');
    if (dotParts.length === 2 && dotParts[1].length <= 2) {
      wholePart = dotParts[0].replace(/\D/g, '');
      decimalPart = dotParts[1].replace(/\D/g, '').padEnd(2, '0').slice(0, 2);
    } else {
      wholePart = clean.replace(/\D/g, '');
      decimalPart = '00';
    }
  } else {
    wholePart = clean.replace(/\D/g, '');
    decimalPart = '00';
  }

  const wholeNum = parseInt(wholePart || '0', 10);
  const decNum = parseInt(decimalPart || '0', 10);
  const totalCents = wholeNum * 100 + decNum;

  return isNegative ? -totalCents : totalCents;
}

/**
 * Formats a live input value string to 'xxx.xxx,00' format on blur or change.
 */
export function formatCurrencyInput(val: string): string {
  if (!val.trim()) return '';
  const cents = parseRupiahToCents(val);
  return formatNumberIDR(cents);
}
