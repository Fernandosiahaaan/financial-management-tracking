/**
 * Quick PIN Authentication Utility for Mobile & Web
 * Allows users to set up a 6-digit PIN for rapid authentication on their devices.
 */

const PIN_HASH_KEY = 'fintrack_pin_hash';
const PIN_EMAIL_KEY = 'fintrack_pin_email';
const PIN_TOKEN_KEY = 'fintrack_pin_token';
const REMEMBERED_EMAIL_KEY = 'fintrack_remembered_email';

export function getRememberedEmail(): string {
  try {
    return (
      localStorage.getItem(REMEMBERED_EMAIL_KEY) ||
      localStorage.getItem(PIN_EMAIL_KEY) ||
      ''
    );
  } catch {
    return '';
  }
}

export function setRememberedEmail(email: string): void {
  try {
    localStorage.setItem(REMEMBERED_EMAIL_KEY, email);
    localStorage.setItem(PIN_EMAIL_KEY, email);
  } catch {
    // Ignore storage errors
  }
}

export function clearRememberedEmail(): void {
  try {
    localStorage.removeItem(REMEMBERED_EMAIL_KEY);
    localStorage.removeItem(PIN_EMAIL_KEY);
  } catch {
    // Ignore storage errors
  }
}

/**
 * Hashes a 6-digit PIN using SHA-256 with salt
 */
export async function hashPin(pin: string): Promise<string> {
  const encoder = new TextEncoder();
  const data = encoder.encode(`fintrack_pin_salt_v1_${pin}`);
  
  if (typeof crypto !== 'undefined' && crypto.subtle) {
    const hashBuffer = await crypto.subtle.digest('SHA-256', data);
    const hashArray = Array.from(new Uint8Array(hashBuffer));
    return hashArray.map((b) => b.toString(16).padStart(2, '0')).join('');
  }
  
  // Simple fallback hash for environments without crypto.subtle
  let hash = 0;
  const str = `fintrack_pin_salt_v1_${pin}`;
  for (let i = 0; i < str.length; i++) {
    const char = str.charCodeAt(i);
    hash = (hash << 5) - hash + char;
    hash |= 0;
  }
  return hash.toString(16);
}

/**
 * Checks if a Quick PIN or remembered session has been configured on this device
 */
export function isPinConfigured(): boolean {
  try {
    return Boolean(getRememberedEmail());
  } catch {
    return false;
  }
}

/**
 * Returns the email of the user associated with the configured PIN
 */
export function getPinUserEmail(): string {
  return getRememberedEmail();
}

/**
 * Sets up a new 6-digit PIN for the current authenticated user
 */
export async function setupPin(pin: string, email: string, token: string): Promise<boolean> {
  if (!pin || pin.length !== 6 || !/^\d{6}$/.test(pin)) {
    throw new Error('PIN harus berupa 6 digit angka');
  }

  const hashed = await hashPin(pin);
  try {
    localStorage.setItem(PIN_HASH_KEY, hashed);
    localStorage.setItem(PIN_EMAIL_KEY, email);
    localStorage.setItem(PIN_TOKEN_KEY, token);
    return true;
  } catch {
    return false;
  }
}

/**
 * Verifies a PIN against the stored hash
 * Returns the stored token if valid, or null if invalid
 */
export async function verifyPin(pin: string): Promise<string | null> {
  if (!pin || pin.length !== 6) return null;

  try {
    const storedHash = localStorage.getItem(PIN_HASH_KEY);
    const storedToken = localStorage.getItem(PIN_TOKEN_KEY);

    if (!storedHash || !storedToken) return null;

    const inputHash = await hashPin(pin);
    if (inputHash === storedHash) {
      return storedToken;
    }
    return null;
  } catch {
    return null;
  }
}

/**
 * Removes the configured PIN from this device
 */
export function removePin(): void {
  try {
    localStorage.removeItem(PIN_HASH_KEY);
    localStorage.removeItem(PIN_EMAIL_KEY);
    localStorage.removeItem(PIN_TOKEN_KEY);
  } catch {
    // Ignore storage errors
  }
}
