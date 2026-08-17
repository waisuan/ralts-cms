// Authentication utilities
import { clearSessionAuthKeys } from './tokens';

/**
 * Dispatched on window when the API rejects a request because the session is
 * over. AuthProvider listens for it and forgets the signed-in user.
 */
export const SESSION_EXPIRED_EVENT = 'ralts:session-expired';

/**
 * Ends the session: forgets the stored user and tokens, then asks the app to
 * show the sign-in screen.
 *
 * This deliberately leaves window.location alone. Assigning a location is a
 * full page load, and when the target is the page that is still loading it
 * aborts that page's own script requests, so the user is left staring at
 * "Loading..." forever. Keeping the URL also means signing back in returns them
 * to the page they were on.
 */
export function endSession(): void {
  if (typeof window === 'undefined') {
    return;
  }
  localStorage.removeItem('ralts_user');
  clearSessionAuthKeys();
  window.dispatchEvent(new Event(SESSION_EXPIRED_EVENT));
}

/**
 * Checks if the current error is a 401 authentication error
 */
export function isAuthError(error: unknown): boolean {
  if (error && typeof error === 'object' && 'status' in error) {
    return (error as { status: number }).status === 401;
  }
  return false;
} 